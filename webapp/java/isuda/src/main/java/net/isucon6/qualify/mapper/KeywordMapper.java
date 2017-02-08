package net.isucon6.qualify.mapper;

import net.isucon6.qualify.domain.Keyword;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

@Mapper
@Repository("keywordMapper")
public interface KeywordMapper {
    void insert(Keyword params);
}
