package net.isucon6.qualify.mapper;

import net.isucon6.qualify.domain.Entry;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

@Mapper
@Repository("entryMapper")
public interface EntryMapper {
    void deleteAll();
    Entry findById(Long id);
}
