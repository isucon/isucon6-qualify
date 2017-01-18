package net.isucon6.qualify.service;

import java.util.List;
import java.util.stream.Collectors;

import net.isucon6.qualify.response.StarResponse;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestOperations;

@Service
public class StarService {

    @Autowired
    private RestOperations isutarRestOperation;

    public List<String> fetch(String keyword) {
        return isutarRestOperation.getForObject("/stars?keyword=" + keyword, StarResponse.class).getStars()
                .stream()
                .map(StarResponse.Star::getUserName)
                .collect(Collectors.toList());
    }

    public void initialize() {
        isutarRestOperation.getForObject("/initialize", String.class);
    }
}
