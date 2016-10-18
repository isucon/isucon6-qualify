import javax.servlet.ServletContext
import skinny.micro.LifeCycle

class Bootstrap extends LifeCycle {
  override def init(ctx: ServletContext): Unit = {
    isutar.Web.setup()
    ctx.mount(isutar.Web, "/*")
  }

  override def destroy(ctx: ServletContext): Unit = {
    isutar.Web.shutdown()
  }
}
