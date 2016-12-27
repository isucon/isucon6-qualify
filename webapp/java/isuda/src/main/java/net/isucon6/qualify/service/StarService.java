package net.isucon6.qualify.service;

import java.util.Arrays;
import java.util.List;

import org.springframework.stereotype.Service;

@Service
public class StarService {
    public List<String> fetch(String keyword) {
        // TODO: API叩く
        return Arrays.asList("blackawa");
    }
}
