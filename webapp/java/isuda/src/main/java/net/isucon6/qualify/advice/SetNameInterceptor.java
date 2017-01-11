package net.isucon6.qualify.advice;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import net.isucon6.qualify.exception.ForbiddenException;
import net.isucon6.qualify.service.UserService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.ModelAndView;
import org.springframework.web.servlet.handler.HandlerInterceptorAdapter;

@Component
public class SetNameInterceptor extends HandlerInterceptorAdapter {
    @Autowired
    private UserService userService;

    private static final Logger log = LoggerFactory.getLogger(SetNameInterceptor.class);

    private static final String USERNAME_KEY = "userName";

    @Override
    public boolean preHandle(
            HttpServletRequest req, HttpServletResponse res, Object handler
    ) throws Exception {
        if (!(handler instanceof HandlerMethod)) return true;
        HandlerMethod handlerMethod = (HandlerMethod) handler;
        if (!handlerMethod.getMethod().isAnnotationPresent(SetName.class)) return true;
        // @SetName がついているエンドポイントでだけ処理を実行する.
        Object userId = req.getSession().getAttribute("userId");

        if (userId == null) return true;
        req.setAttribute("userId", userId);
        String userName = userService.findNameById(Long.valueOf(String.valueOf(userId)));

        if (userName == null) throw new ForbiddenException();
        req.setAttribute(USERNAME_KEY, userName);
        return true;
    }

    @Override
    public void postHandle(
            HttpServletRequest req, HttpServletResponse res, Object handler, ModelAndView mav
    ) throws Exception {
        mav.addObject(USERNAME_KEY, req.getAttribute(USERNAME_KEY));
    }
}
