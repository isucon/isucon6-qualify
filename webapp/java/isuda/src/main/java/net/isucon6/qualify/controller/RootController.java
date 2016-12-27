package net.isucon6.qualify.controller;

import java.util.HashMap;

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
    private final int PER_PAGE = 10;
    @RequestMapping(value = "/")
    public ModelAndView getIndex(@RequestParam(defaultValue = "1") int page) {
        ModelAndView mav = new ModelAndView();
        // TODO: Controller共通の処理として、Sessionがあればrequestに毎回ユーザー名を突っ込んでおく
        mav.addObject("request", new HashMap<String, String>() {{
            put("userName", "username1");
        }});

        mav.addObject("entries", entryService.findEntries(PER_PAGE, page));
        mav.addObject("page", page);
        mav.addObject("paging", pagingService.fetchEntryPage(PER_PAGE, page));

        mav.setViewName("index");
        return mav;
    }
}
