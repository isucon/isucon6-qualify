package net.isucon6.qualify.controller;

import java.util.HashMap;
import java.util.Map;

import net.isucon6.qualify.mapper.EntryMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class InitializeController {
    @Autowired
    private EntryMapper entryMapper;

    @RequestMapping(value = "/initialize")
    public Map<String, String> getInitialize() {
        entryMapper.initialize();
        // TODO: backendTemplateからisutarのinitializeを呼び出す
        return new HashMap<String, String>() {{
            put("result", "ok");
        }};
    }
}
