package net.isucon6.qualify.service;

import java.util.List;

import net.isucon6.qualify.domain.Star;
import net.isucon6.qualify.exception.NotFoundException;
import net.isucon6.qualify.mapper.StarMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class StarService {
    @Autowired
    private StarMapper starMapper;
    @Autowired
    private KeywordService keywordService;

    public void initialize() {
        starMapper.initialize();
    }

    public List<Star> findByKeyword(String keyword) {
        return starMapper.findByKeyword(keyword);
    }

    public void insert(String keyword, String userName) {
        if (!keywordService.exists(keyword)) {
            throw new NotFoundException();
        }
        starMapper.insert(new Star(keyword, userName));
    }
}
