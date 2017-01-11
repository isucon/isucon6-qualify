package net.isucon6.qualify.controller;

import java.util.List;

import net.isucon6.qualify.domain.Star;
import net.isucon6.qualify.service.StarService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class StarController {
    @Autowired
    private StarService starService;

    @RequestMapping(value = "/stars")
    public List<Star> findByKeyword(@RequestParam("keyword") String keyword) {
        return starService.findByKeyword(keyword);
    }
}
