package net.isucon6.qualify.controller;

import java.util.HashMap;

import net.isucon6.qualify.advice.SetName;
import net.isucon6.qualify.service.EntryService;
import net.isucon6.qualify.service.PagingService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.servlet.ModelAndView;

@Controller
public class RootController {
    @Autowired
    private EntryService entryService;
    @Autowired
    private PagingService pagingService;

    @SetName
    @RequestMapping(value = "/")
    public ModelAndView getIndex(@RequestParam(defaultValue = "1") int page) {
        ModelAndView mav = new ModelAndView();
        mav.addObject("request", new HashMap<String, String>() {{
            put("userName", "username1");
        }});

        int perPage = 10;
        mav.addObject("entries", entryService.findHtmlEntries(perPage, page));
        mav.addObject("page", page);
        mav.addObject("paging", pagingService.fetchEntryPage(perPage, page));

        mav.setViewName("index");
        return mav;
    }
}
