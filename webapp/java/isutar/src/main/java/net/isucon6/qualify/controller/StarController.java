package net.isucon6.qualify.controller;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.isucon6.qualify.domain.Star;
import net.isucon6.qualify.form.StarForm;
import net.isucon6.qualify.service.StarService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class StarController {
    @Autowired
    private StarService starService;

    @RequestMapping(value = "/stars", method = RequestMethod.GET)
    public List<Star> findByKeyword(@RequestParam("keyword") String keyword) {
        return starService.findByKeyword(keyword);
    }

    @RequestMapping(value = "/stars", method = RequestMethod.POST)
    public Map<String, String> insert(
            @RequestBody(required = false) StarForm json,
            @ModelAttribute StarForm params
    ) {
        String keyword = StringUtils.isEmpty(params.getKeyword()) ? json.getKeyword() : params.getKeyword();
        String userName = StringUtils.isEmpty(params.getUser()) ? json.getUser() : params.getUser();
        starService.insert(keyword, userName);
        return new HashMap<String, String>() {{
            put("result", "ok");
        }};
    }
}
