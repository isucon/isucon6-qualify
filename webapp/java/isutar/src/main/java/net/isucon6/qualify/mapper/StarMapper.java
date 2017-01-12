package net.isucon6.qualify.mapper;

import java.util.List;

import net.isucon6.qualify.domain.Star;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

@Mapper
@Repository
public interface StarMapper {
    void initialize();
    List<Star> findByKeyword(String keyword);
    void insert(Star star);
}
