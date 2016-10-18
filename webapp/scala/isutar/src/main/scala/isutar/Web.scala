package isutar

import com.typesafe.config.ConfigFactory
import org.joda.time.LocalDateTime
import org.json4s.{CustomSerializer, DefaultFormats, Formats, JString}
import scala.util.Try
import scalikejdbc.{DB, WrappedResultSet}
import scalikejdbc.config.DBs
import scalikejdbc.interpolation.Implicits._
import skinny.micro.{NotFound, Ok, WebApp}
import skinny.micro.contrib.json4s.JSONSupport
import skinny.http

object Web extends WebApp with JSONSupport {
  import Util.{HttpResponseOps, StringConversion}

  get("/initialize") {
    DB.autoCommit { implicit session =>
      sql"TRUNCATE star".update.apply()
    }
    Ok(toJSONString(Map("result" -> "ok")))
  }

  get("/stars") {
    val keyword = params.get("keyword").getOrElse("")
    val stars = DB.readOnly { implicit session =>
      sql"""
        SELECT * FROM star WHERE keyword = $keyword
      """.map(asStar).list.apply()
    }
    Ok(toJSONString(Map("stars" -> stars)))
  }

  post("/stars") {
    val keyword = params.get("keyword").getOrElse("")
    val origin = appConfig.getString("origin.isuda")
    val url = s"$origin/keyword/${keyword.uriEncoded}"
    val res = http.HTTP.get(url)

    if (res.isSuccess) {
      val user = params.get("user").getOrElse("")
      DB.autoCommit { implicit session =>
        sql"""
        INSERT INTO star (keyword, user_name, created_at)
        VALUES ($keyword, $user, NOW())
      """.update.apply()
      }

      Ok(toJSONString(Map("result" -> "ok")))
    } else NotFound()
  }

  def asStar(rs: WrappedResultSet): Model.Star = Model.Star(
    rs.get[Long]("id"),
    rs.get[String]("keyword"),
    rs.get[String]("user_name"),
    rs.get[LocalDateTime]("created_at")
  )

  def setup(): Unit = {
    DBs.setupAll
  }

  override def shutdown(): Unit = {
    super.shutdown()
    DBs.closeAll
  }

  lazy val appConfig = ConfigFactory.load

  override protected implicit val jsonFormats: Formats =
    DefaultFormats ++ Seq(LocalDateTimeSerializer)
}

object Model {
  case class Star(
    id: Long,
    keyword: String,
    userName: String,
    createdAt: LocalDateTime
  )
}

class LocalDateTimeSerializer
    extends CustomSerializer[LocalDateTime](format => ({
      case JString(value) => new LocalDateTime(value)
    }, {
      case dt: LocalDateTime =>
        JString(dt.toString.dropRight(4).replaceAll("T", " "))
    }))
object LocalDateTimeSerializer extends LocalDateTimeSerializer

object Util {
  implicit class StringConversion(private val str: String) extends AnyVal {
    def uriEncoded(): String = Try {
      val s = java.net.URLEncoder.encode(str, "UTF-8")
      Seq(
        // RFC3986
        "+" -> "%20",
        "%7E" -> "~"
      ).foldLeft(s) { (s, subst) => s.replaceAllLiterally(subst._1, subst._2) }
    } getOrElse str
  }

  implicit class HttpResponseOps(private val res: http.Response) extends AnyVal {
    def isSuccess: Boolean = res.status >= 200 && res.status < 300
  }
}
