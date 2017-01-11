package net.isucon6.qualify.service;

import java.util.List;

import net.isucon6.qualify.domain.Star;
import net.isucon6.qualify.mapper.StarMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class StarService {
    @Autowired
    private StarMapper starMapper;

    public void initialize() {
        starMapper.initialize();
    }

    public List<Star> findByKeyword(String keyword) {
        return starMapper.findByKeyword(keyword);
    }
}
