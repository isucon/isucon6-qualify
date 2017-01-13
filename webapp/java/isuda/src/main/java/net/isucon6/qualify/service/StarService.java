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
    private RestOperations isutarRestOperation;

    public List<String> fetch(String keyword) {
        return (List<String>) isutarRestOperation.getForObject("/stars?keyword=" + keyword, List.class)
                .stream()
                .map(star -> ((Map) star).get("userName"))
                .collect(Collectors.toList());
    }

    public void initialize() {
        isutarRestOperation.getForObject("/initialize", String.class);
    }
}
