package net.isucon6.qualify.advice;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import net.isucon6.qualify.exception.ForbiddenException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.handler.HandlerInterceptorAdapter;

@Component
public class AuthenticateInterceptor extends HandlerInterceptorAdapter {
    private static final Logger log = LoggerFactory.getLogger(AuthenticateInterceptor.class);

    @Override
    public boolean preHandle(
            HttpServletRequest req, HttpServletResponse res, Object handler
    ) throws Exception {
        if (!(handler instanceof HandlerMethod)) return true;
        HandlerMethod handlerMethod = (HandlerMethod) handler;
        if (!handlerMethod.getMethod().isAnnotationPresent(Authenticate.class)) return true;

        // @Authenticate がついているエンドポイントでだけ処理を実行する.
        log.info("check authentication");
        if (req.getAttribute("userId") == null) throw new ForbiddenException();
        return true;
    }
}
