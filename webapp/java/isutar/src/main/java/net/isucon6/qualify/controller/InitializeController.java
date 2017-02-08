package net.isucon6.qualify.controller;

import java.util.HashMap;
import java.util.Map;

import net.isucon6.qualify.service.StarService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class InitializeController {
    @Autowired
    private StarService starService;

    @RequestMapping(value = "/initialize")
    public Map<String, String> initialize() {
        starService.initialize();
        return new HashMap<String, String>() {{
            put("status", "ok");
        }};
    }
}
