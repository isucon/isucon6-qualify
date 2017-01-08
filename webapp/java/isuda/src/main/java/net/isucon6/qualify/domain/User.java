package net.isucon6.qualify.domain;

import java.util.Date;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.apache.commons.codec.digest.DigestUtils;
import org.apache.commons.lang3.RandomStringUtils;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class User {
    private long id;
    private String name;
    private String salt;
    private String password;
    private Date createdAt;

    public User(String name, String password) {
        this.name = name;
        this.salt = RandomStringUtils.random(20, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789");
        this.password = DigestUtils.sha1Hex(salt + password);
    }
}
