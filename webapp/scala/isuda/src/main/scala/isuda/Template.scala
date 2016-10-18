package isuda

import skinny.micro.context.SkinnyContext
import skinny.micro.contrib.ScalateSupport
import skinny.micro.response.{ActionResult, Ok}

trait Template extends ScalateSupport {
  def render(path: String, attributes: (String, Any)*)(implicit
    ctx: SkinnyContext
  ): ActionResult = {
    val html = mustache(path, attributes: _*)
    Ok(html, Map("Content-Type" -> "text/html"))
  }
}
