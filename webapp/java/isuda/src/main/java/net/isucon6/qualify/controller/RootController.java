package net.isucon6.qualify.controller;

import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.List;

import net.isucon6.qualify.domain.Entry;
import net.isucon6.qualify.domain.Star;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.servlet.ModelAndView;

@Controller
public class RootController {
    @RequestMapping(value = "/")
    public ModelAndView getIndex() {
        ModelAndView mav = new ModelAndView();
        mav.addObject("request", new HashMap<String, String>() {{
            put("userName", "username1");
        }});
        mav.addObject("page", 1);
        mav.addObject("lastPage", 5);
        mav.addObject("pages", new int[]{1, 2, 3, 4, 5});
        List<Entry> entries = new ArrayList<Entry>() {{
            add(new Entry(1L, 1L, "キーワード1", "説明1", new Date(), new Date(), new ArrayList<Star>() {{
                add(new Star(1L, "キーワード1", "blackawa", new Date()));
            }}));
            add(new Entry(2L, 1L, "キーワード2", "説明2", new Date(), new Date(), new ArrayList<>()));
        }};
        mav.addObject("entries", entries);
        mav.setViewName("index");
        return mav;
    }
}
