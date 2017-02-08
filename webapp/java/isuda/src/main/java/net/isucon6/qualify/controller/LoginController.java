package net.isucon6.qualify.controller;

import javax.servlet.http.HttpSession;

import net.isucon6.qualify.advice.SetName;
import net.isucon6.qualify.domain.User;
import net.isucon6.qualify.exception.ForbiddenException;
import net.isucon6.qualify.form.AuthenticateForm;
import net.isucon6.qualify.mapper.UserMapper;
import org.apache.commons.codec.digest.DigestUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.servlet.ModelAndView;

@Controller
public class LoginController {
    @Autowired
    private UserMapper userMapper;

    @SetName
    @RequestMapping(value = "/login")
    public ModelAndView index() {
        ModelAndView mav = new ModelAndView();
        mav.addObject("action", "login");
        mav.addObject("authenticate", new AuthenticateForm());
        mav.setViewName("authenticate");
        return mav;
    }

    @RequestMapping(value = "/login", method = RequestMethod.POST)
    public ModelAndView login(@ModelAttribute AuthenticateForm form, HttpSession session) {
        User user = userMapper.findByName(form.getName());
        if (user == null || !user.getPassword().equals(DigestUtils.sha1Hex(user.getSalt() + form.getPassword()))) {
            throw new ForbiddenException();
        }
        ModelAndView mav = new ModelAndView();
        mav.setViewName("redirect:/");
        session.setAttribute("userId", user.getId());
        return mav;
    }
}
