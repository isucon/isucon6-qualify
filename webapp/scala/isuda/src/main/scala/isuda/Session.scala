package isuda

import java.io.{
  ByteArrayInputStream,
  ByteArrayOutputStream,
  ObjectInputStream,
  ObjectOutputStream
}
import java.util.Base64
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec
import javax.servlet.http.{Cookie => HttpCookie}
import scala.collection.mutable.{HashMap => MutableMap}
import scala.util.control.Exception.allCatch
import skinny.micro.SkinnyMicroFilter
import skinny.micro.context.SkinnyContext

trait Session {
  def sessionStore(implicit ctx: SkinnyContext): Session.Store
}

object Session {
  trait Store {
    def put(key: String, value: String): Unit
    def get(key: String): Option[String]
    def clear(): Unit
  }

  trait Cookie extends Session {
    self: SkinnyMicroFilter =>

    val sessionKey: String = "skinny_micro_session"
    def sessionSecretKey: String

    def sessionStore(implicit ctx: SkinnyContext): Cookie.Store = {
      val store =
        request(ctx).getAttribute(Cookie.SessionAttr).asInstanceOf[Cookie.Store]
      Option(store).getOrElse(Cookie.Store())
    }

    before() {
      val req = request
      val store =
        req.cookie(sessionKey).flatMap(deserialize(_)).getOrElse(Cookie.Store())
      req.setAttribute(Cookie.SessionAttr, store)
    }

    after() {
      val res = response
      val cookie = new HttpCookie(sessionKey, serialize(sessionStore))
      cookie.setPath("/")
      res.addCookie(cookie)
    }

    private def serialize(store: Cookie.Store): String = {
      val base64 = store.serialized
      Seq(base64, sign(sessionSecretKey, base64)).mkString(":")
    }

    private def deserialize(cookie: String): Option[Cookie.Store] =
      cookie.split(":").toSeq match {
        case base64 +: sig +: Nil if sign(sessionSecretKey, base64) == sig =>
          Some(Cookie.Store.deserialize(base64))
        case _ => None
      }

    final val HmacSHA1 = "HmacSHA1"

    private def sign(secret: String, data: String): String = {
      val mac = Mac.getInstance(HmacSHA1)
      mac.init(new SecretKeySpec(secret.getBytes("UTF-8"), HmacSHA1))
      mac.doFinal(data.getBytes("UTF-8")).map("%02X" format _).mkString
    }
  }
  object Cookie {
    final val SessionAttr: String = Seq(
      classOf[Cookie].getCanonicalName,
      "session"
    ).mkString("_")

    class Store extends Session.Store {
      private val map: MutableMap[String, String] = new MutableMap
      def m = map
      def put(key: String, value: String): Unit = map.put(key, value)
      def get(key: String): Option[String] = map.get(key)
      def clear(): Unit = map.clear()
      def serialized: String = {
        val baos = new ByteArrayOutputStream
        val oos = new ObjectOutputStream(baos)
        try oos.writeObject(map) finally oos.close
        Base64.getEncoder.encodeToString(baos.toByteArray)
      }
    }
    object Store {
      def apply(values: (String, String)*): Store = {
        val store = new Store
        values.foreach { case (key, value) =>
          store.put(key, value)
        }
        store
      }
      def deserialize(base64: String): Store =
        allCatch.opt {
          val bytes = Base64.getDecoder.decode(base64)
          val bais = new ByteArrayInputStream(bytes)
          val ois = new ObjectInputStream(bais)
          val map =
            try ois.readObject().asInstanceOf[MutableMap[String, String]]
            finally ois.close
          Store(map.toSeq: _*)
        }.getOrElse(Store())
    }
  }
}
