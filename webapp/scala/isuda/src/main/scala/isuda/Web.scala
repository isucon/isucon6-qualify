package isuda

import com.typesafe.config.ConfigFactory
import java.security.{MessageDigest, SecureRandom}
import java.util.regex.Pattern
import org.joda.time.LocalDateTime
import scala.util.{Random, Try}
import scala.util.control.Exception.catching
import scalikejdbc.{DB, WrappedResultSet}
import scalikejdbc.config.DBs
import scalikejdbc.interpolation.Implicits._
import skinny.micro.{ActionResult, BadRequest, Forbidden, NotFound, Ok, WebApp}
import skinny.micro.contrib.json4s.JSONSupport
import skinny.http

object Web extends WebApp
    with Static
    with Session.Cookie
    with Template
    with JSONSupport {
  import Util.{sha1Hex, StringConversion}

  final val PerPage = 10

  get("/initialize") {
    DB.autoCommit { implicit session =>
      sql"DELETE FROM entry WHERE id > 7101".update.apply()
    }

    val origin = appConfig.getString("origin.isutar")
    val url = s"$origin/initialize"
    http.HTTP.get(url)

    Ok(toJSONString(Map("result" -> "ok")))
  }

  get("/")(withUserName { maybeUserName =>
    val page = params.get("page").flatMap(_.asInt).getOrElse(1)
    val offset = PerPage * (page - 1)
    val entries = DB.readOnly { implicit session =>
      sql"""
        SELECT * FROM entry
        ORDER BY updated_at DESC
        LIMIT $PerPage
        OFFSET $offset
      """.map(asEntry).list.apply()
    }

    val Some(totalEntries) = DB.readOnly { implicit session =>
      sql"SELECT COUNT(*) FROM entry".map(_.int(1)).single.apply()
    }
    val lastPage = totalEntries / PerPage
    val pages = Seq(1, page - 5).max to Seq(lastPage, page + 5).min

    render("index",
      "user" -> maybeUserName,
      "entries" -> entries.map(e => Map("entry" -> e.toHash)),
      "pages" -> pages.map(p => Map(
        "page" -> p,
        "active" -> (p == page)
      )),
      "prev" -> Option(page - 1).filter(_ > 0),
      "next" -> Option(page + 1).filter(_ < lastPage)
    )
  })

  get("robots.txt") {
    NotFound()
  }

  post("/keyword")(withAuthorizedUserId { userId =>
    val description = params.get("description").getOrElse("")
    val result = for {
      keyword <- params.get("keyword").filter(_.nonEmpty).toRight(BadRequest()).right
      _ <- Either.cond(
        isSpamContents(description) || isSpamContents(keyword),
        BadRequest("SPAM!"),
        Unit
      ).swap.right
    } yield {
      DB.autoCommit { implicit session =>
        sql"""
          INSERT INTO entry (author_id, keyword, description, created_at, updated_at)
          VALUES ($userId, $keyword, $description, NOW(), NOW())
          ON DUPLICATE KEY UPDATE
            author_id = VALUES(author_id),
            keyword = VALUES(keyword),
            description = VALUES(description),
            created_at = created_at,
            updated_at = VALUES(updated_at)
        """.update.apply()
      }
      redirect(uriFor("/"))
    }
    result.merge
  })

  get("/register")(withUserName { maybeUserName =>
    render("authenticate",
      "user" -> maybeUserName,
      "Action" -> "Register",
      "action" -> "register"
    )
  })

  post("/register") {
    def register(user: String, pass: String): Long = {
      val salt = new Random(new SecureRandom).alphanumeric.take(20).mkString
      val sha1 = sha1Hex(salt + pass)
      val Some(id) = DB.autoCommit { implicit session =>
        sql"""
          INSERT INTO user (name, salt, password, created_at)
          VALUES ($user, $salt, $sha1, NOW())
        """.update.apply()
        sql"SELECT LAST_INSERT_ID()".map(_.long(1)).single.apply()
      }
      id
    }

    val maybeId = for {
      name <- params.get("name").filter(_.nonEmpty)
      pass <- params.get("password").filter(_.nonEmpty)
    } yield register(name, pass)
    maybeId.fold(BadRequest()) { id =>
      sessionStore.put("userId", id.toString)
      redirect(uriFor("/"))
    }
  }

  get("/login")(withUserName { maybeUserName =>
    render("authenticate",
      "user" -> maybeUserName,
      "Action" -> "Login",
      "action" -> "login"
    )
  })

  post("/login") {
    def loadUser(name: String, pass: String): Option[Model.User] =
      DB.readOnly { implicit session =>
        sql"""
          SELECT * FROM user
          WHERE name = $name
        """.map(asUser).single.apply()
      }.filter(user => user.password == sha1Hex(user.salt + pass))

    val maybeUser = for {
      name <- params.get("name").filter(_.nonEmpty)
      pass <- params.get("password").filter(_.nonEmpty)
      user <- loadUser(name, pass)
    } yield user
    maybeUser.fold(Forbidden()) { user =>
      sessionStore.put("userId", user.id.toString)
      redirect(uriFor("/"))
    }
  }

  get("/logout") {
    sessionStore.clear()
    redirect(uriFor("/"))
  }

  get("/keyword/:keyword")(withUserName { maybeUserName =>
    val result = for {
      keyword <- params.get("keyword").toRight(BadRequest()).right
      entry <- DB.readOnly { implicit session =>
        sql"""
          SELECT * FROM entry
          WHERE keyword = $keyword
        """.map(asEntry).single.apply()
      }.toRight(NotFound()).right
    } yield render("keyword",
      "user" -> maybeUserName,
      "entry" -> entry.toHash
    )
    result.merge
  })

  post("/keyword/:keyword")(withAuthorizedUserId { _ =>
    val result = for {
      keyword <- params.get("keyword").toRight(BadRequest()).right
      _ <- params.get("delete").toRight(BadRequest()).right
      _ <- DB.readOnly { implicit session =>
        sql"""
          SELECT id FROM entry
          WHERE keyword = $keyword
        """.map(_.long(1)).single.apply()
      }.toRight(NotFound()).right
    } yield {
      DB.autoCommit { implicit session =>
        sql"""
          DELETE FROM entry
          WHERE keyword = $keyword
        """.update.apply()
      }
      redirect(uriFor("/"))
    }
    result.merge
  })

  private def withLogin(
    block: Option[Model.Login] => ActionResult
  ): ActionResult = sessionStore.get("userId").flatMap(_.asLong).map { userId =>
    DB.readOnly { implicit session =>
      val maybeLogin =
        sql"""
          SELECT name FROM user
          WHERE id = $userId
        """.map(rs => Model.Login(userId, rs.string(1))).single.apply()
      maybeLogin match {
        case Some(login) =>  block(Some(login))
        case None => Forbidden()
      }
    }
  }.getOrElse(block(None))

  def withUserName(block: Option[String] => ActionResult): ActionResult =
    withLogin(login => block(login.map(_.userName)))

  def withAuthorizedUserId(block: Long => ActionResult): ActionResult =
    withLogin {
      case Some(login) => block(login.userId)
      case None => Forbidden()
    }

  def htmlify(content: String): String = {
    val entries = DB.readOnly { implicit session =>
      sql"""
        SELECT * FROM entry
        ORDER BY CHARACTER_LENGTH(keyword) DESC
      """.map(asEntry).list.apply()
    }
    val regex =
      entries.map(e => Pattern.quote(e.keyword)).mkString("(", "|", ")").r
    val hashBuilder = Map.newBuilder[String, String]
    val escaped = regex.replaceAllIn(content, m => {
      val kw = m.group(1)
      val hash = s"isuda_${sha1Hex(kw)}"
      hashBuilder += kw -> hash
      hash
    }).htmlEscaped
    hashBuilder.result.foldLeft(escaped) { case (content, (kw, hash)) =>
      val url = s"/keyword/${kw.uriEncoded}"
      val link = s"""<a href="$url">${kw.htmlEscaped}</a>"""
      content.replaceAllLiterally(hash, link)
    }.replaceAllLiterally("\n", "<br />\n")
  }

  def loadStars(keyword: String): Seq[Model.Star] = {
    val origin = appConfig.getString("origin.isutar")
    val url = s"$origin/stars"
    val res = http.HTTP.get(url, "keyword" -> keyword)
    val maybeStars = fromJSONString[Model.Stars](res.textBody).toOption
    maybeStars.fold(Seq.empty[Model.Star])(_.stars)
  }

  def isSpamContents(content: String): Boolean = {
    val origin = appConfig.getString("origin.isupam")
    val url = s"$origin/"
    val res = http.HTTP.post(url, "content" -> content)
    val maybeValidity =
      fromJSONString[Model.ContentValidity](res.textBody).toOption
    !maybeValidity.fold(false)(_.valid)
  }

  def uriFor(path: String): String =
    if (path.endsWith("/")) s"${fullUrl(path)}/"
    else fullUrl(path)

  implicit class EntryOps(private val entry: Model.Entry) extends AnyVal {
    def toHash: Map[String, Any] =
      toJSON(entry).extract[Map[String, Any]] ++ Map(
        "url" -> s"/keyword/${entry.keyword.uriEncoded}",
        "html" -> htmlify(entry.description),
        "stars" -> loadStars(entry.keyword).map(s => Map("star" -> s))
      )
  }

  def asEntry(rs: WrappedResultSet): Model.Entry = Model.Entry(
    rs.get[Long]("id"),
    rs.get[Long]("author_id"),
    rs.get[String]("keyword"),
    rs.get[String]("description"),
    rs.get[LocalDateTime]("updated_at"),
    rs.get[LocalDateTime]("created_at")
  )

  def asUser(rs: WrappedResultSet): Model.User = Model.User(
    rs.get[Long]("id"),
    rs.get[String]("name"),
    rs.get[String]("password"),
    rs.get[String]("salt"),
    rs.get[LocalDateTime]("created_at")
  )

  override val sessionKey: String = "isuda_session"
  val sessionSecretKey: String = "tonymoris"

  def setup(): Unit = {
    DBs.setupAll
  }

  override def shutdown(): Unit = {
    super.shutdown()
    DBs.closeAll
  }

  lazy val appConfig = ConfigFactory.load
}

object Model {
  case class Entry(
    id: Long,
    authorId: Long,
    keyword: String,
    description: String,
    updatedAt: LocalDateTime,
    createdAt: LocalDateTime
  )

  case class User(
    id: Long,
    name: String,
    password: String,
    salt: String,
    createdAt: LocalDateTime
  )

  case class Login(
    userId: Long,
    userName: String
  )

  case class Star(
    id: Long,
    keyword: String,
    userName: String
  )

  case class Stars(stars: Seq[Star])

  case class ContentValidity(valid: Boolean)
}

object Util {
  def sha1Hex(str: String): String = {
    val md = MessageDigest.getInstance("SHA-1")
    md.digest(str.getBytes("UTF-8")).map("%02x" format _).mkString
  }

  implicit class StringConversion(private val str: String) extends AnyVal {
    def asInt: Option[Int] =
      catching(classOf[NumberFormatException]) opt str.toInt
    def asLong: Option[Long] =
      catching(classOf[NumberFormatException]) opt str.toLong

    def uriEncoded(): String = Try {
      val s = java.net.URLEncoder.encode(str, "UTF-8")
      Seq(
        // RFC3986
        "+" -> "%20",
        "%7E" -> "~"
      ).foldLeft(s) { (s, subst) => s.replaceAllLiterally(subst._1, subst._2) }
    } getOrElse str

    def htmlEscaped(): String = Seq(
      "&" -> "&amp;",
      "<" -> "&lt;",
      ">" -> "&gt;",
      "\"" -> "&quot;"
    ).foldLeft(str) { (s, subst) => s.replaceAllLiterally(subst._1, subst._2) }
  }
}
