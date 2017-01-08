package net.isucon6.qualify.service;

import net.isucon6.qualify.domain.User;
import net.isucon6.qualify.mapper.UserMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class UserService {
    @Autowired
    private UserMapper userMapper;

    public int create(String name, String password) {
        userMapper.create(new User(name, password));
        return userMapper.findLastInsertId();
    }
}
