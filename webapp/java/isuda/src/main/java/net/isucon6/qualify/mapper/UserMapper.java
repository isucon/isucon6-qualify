package net.isucon6.qualify.mapper;

import net.isucon6.qualify.domain.User;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

@Mapper
@Repository
public interface UserMapper {
    User findByName(String username);
    void create(User params);
    int findLastInsertId();
}
