package net.isucon6.qualify.controller;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import javax.servlet.http.HttpServletRequest;

import net.isucon6.qualify.domain.Star;
import net.isucon6.qualify.form.StarForm;
import net.isucon6.qualify.service.StarService;
import org.springframework.beans.MutablePropertyValues;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.validation.BindingResult;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.WebDataBinder;
import org.springframework.web.bind.annotation.InitBinder;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class StarController {
    @Autowired
    private StarService starService;

    @InitBinder("form")
    public void bindStar(WebDataBinder binder, HttpServletRequest request) {
        MutablePropertyValues pvs = new MutablePropertyValues();
        pvs.add("userName", request.getParameter("user_name"));
        binder.bind(pvs);
    }

    @ModelAttribute("form")
    public StarForm newStar() {
        return new StarForm();
    }

    @RequestMapping(value = "/stars")
    public List<Star> findByKeyword(@RequestParam("keyword") String keyword) {
        return starService.findByKeyword(keyword);
    }

    @RequestMapping(value = "/stars", method = RequestMethod.POST)
    public Map<String, String> insert(@Validated @ModelAttribute StarForm form, BindingResult bindingResult) {
        System.out.println(form);
        return new HashMap<String, String>() {{
            put("result", "ok");
        }};
    }
}
