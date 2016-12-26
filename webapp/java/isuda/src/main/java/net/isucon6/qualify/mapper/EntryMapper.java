package net.isucon6.qualify.mapper;

import java.util.List;
import java.util.Map;

import net.isucon6.qualify.domain.Entry;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

@Mapper
@Repository("entryMapper")
public interface EntryMapper {
    void initialize();
    List<Entry> findByPageNum(Map<String, Integer> params);
}
