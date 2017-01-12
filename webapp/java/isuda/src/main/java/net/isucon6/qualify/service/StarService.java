package net.isucon6.qualify.service;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestOperations;

@Service
public class StarService {
    @Autowired
    private RestOperations isutarRestOperations;

    public List<String> fetch(String keyword) {
        // TODO: API叩く
        return (List<String>) isutarRestOperations.getForObject("/stars?keyword=" + keyword, List.class)
                .stream()
                .map(star -> ((Map) star).get("userName"))
                .collect(Collectors.toList());
    }

    public void initialize() {
        isutarRestOperations.getForObject("/initialize", String.class);
    }
}
