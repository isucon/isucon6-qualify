package net.isucon6.qualify.controller;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.isucon6.qualify.domain.Star;
import net.isucon6.qualify.form.StarForm;
import net.isucon6.qualify.service.StarService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class StarController {
    @Autowired
    private StarService starService;

    @RequestMapping(value = "/stars", method = RequestMethod.GET)
    public Map<String, List<Star>> findByKeyword(@RequestParam("keyword") String keyword) {
        return new HashMap<String, List<Star>>() {{
            put("stars", starService.findByKeyword(keyword));
        }};
    }

    @RequestMapping(value = "/stars", method = RequestMethod.POST)
    public Map<String, String> insert(@ModelAttribute StarForm params) {
        String keyword = params.getKeyword();
        String userName = params.getUser();
        starService.insert(keyword, userName);
        return new HashMap<String, String>() {{
            put("result", "ok");
        }};
    }
}
