package net.isucon6.qualify.controller;

import java.util.HashMap;

import net.isucon6.qualify.service.EntryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.servlet.ModelAndView;

@Controller
public class RootController {
    @Autowired
    private EntryService entryService;
    @RequestMapping(value = "/")
    public ModelAndView getIndex(@RequestParam(defaultValue = "1") int page) {
        ModelAndView mav = new ModelAndView();
        // TODO: Aspectで、Sessionがあればrequestに毎回ユーザー名を突っ込んでおく
        mav.addObject("request", new HashMap<String, String>() {{
            put("userName", "username1");
        }});

        // TODO: 現在のページや全体のページ数を出すサービスに切り出す
        int perPage = 10;
        mav.addObject("page", page);
        int[] pages = {1, 2, 3, 4, 5};
        mav.addObject("lastPage", pages.length);
        mav.addObject("pages", pages);

        mav.addObject("entries", entryService.findEntries(perPage, page));

        mav.setViewName("index");
        return mav;
    }
}
