package isuda

import skinny.micro.{ApiFormats, SkinnyMicroFilter, NotFound, Ok}
import scala.util.control.Exception.allCatch

trait Static extends ApiFormats {
  self: SkinnyMicroFilter =>

  addMimeMapping("image/x-icon", "ico")

  get("/(?:(?:css|js|img)/.*|favicon[.]ico)".r) {
    val ext = requestPath.split("[.]").lastOption
    response.contentType = ext.flatMap(formats.get(_))

    val maybeFile =
      allCatch.opt(new java.io.File(getClass.getResource(requestPath).toURI))
    maybeFile.fold(NotFound())(Ok(_))
  }
}
