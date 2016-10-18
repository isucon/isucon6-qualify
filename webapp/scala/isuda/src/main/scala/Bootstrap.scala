import javax.servlet.ServletContext
import skinny.micro.LifeCycle

class Bootstrap extends LifeCycle {
  override def init(ctx: ServletContext): Unit = {
    isuda.Web.setup()
    ctx.mount(isuda.Web, "/*")
  }

  override def destroy(ctx: ServletContext): Unit = {
    isuda.Web.shutdown()
  }
}
