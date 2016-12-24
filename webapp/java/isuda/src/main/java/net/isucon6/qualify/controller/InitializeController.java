package net.isucon6.qualify.controller;

import net.isucon6.qualify.domain.Entry;
import net.isucon6.qualify.mapper.EntryMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class InitializeController {
    @Autowired
    private EntryMapper entryMapper;

    @RequestMapping(value = "/initialize")
    public String getInitialize() {
        return "Initialized";
    }

    @RequestMapping(value = "/demo")
    public Entry getEntry() {
        return entryMapper.findById(1L);
    }
}
