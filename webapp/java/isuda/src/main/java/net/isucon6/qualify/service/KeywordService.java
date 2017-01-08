package net.isucon6.qualify.service;

import net.isucon6.qualify.domain.Keyword;
import net.isucon6.qualify.mapper.KeywordMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class KeywordService {
    @Autowired
    private KeywordMapper keywordMapper;

    public void insert(Keyword keyword) {
        keywordMapper.insert(keyword);
    }
}
