package net.isucon6.qualify.controller;

import java.util.HashMap;
import java.util.Map;

import javax.servlet.http.HttpSession;
import javax.validation.Valid;

import net.isucon6.qualify.advice.SetName;
import net.isucon6.qualify.exception.BadRequestException;
import net.isucon6.qualify.form.AuthenticateForm;
import net.isucon6.qualify.service.UserService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Controller;
import org.springframework.validation.BindingResult;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.servlet.ModelAndView;

@Controller
public class RegisterController {
    @Autowired
    private UserService userService;

    @SetName
    @RequestMapping(value = "/register")
    public ModelAndView show() {
        Map<String, Object> params = new HashMap<>();
        params.put("action", "register");
        params.put("authenticate", new AuthenticateForm());
        return new ModelAndView("authenticate", params);
    }

    @RequestMapping(value = "/register", method = RequestMethod.POST)
    public ModelAndView create(@Valid @ModelAttribute AuthenticateForm form, BindingResult bindingResult, HttpSession session) {
        if (bindingResult.hasErrors()) {
            throw new BadRequestException();
        }
        int userId = userService.create(form.getName(), form.getPassword());
        session.setAttribute("userId", userId);
        return new ModelAndView("redirect:/");
    }
}
