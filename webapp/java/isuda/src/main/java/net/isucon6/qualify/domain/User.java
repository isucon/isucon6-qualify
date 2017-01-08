package net.isucon6.qualify.domain;

import java.util.Date;

import lombok.Data;

@Data
public class User {
    private long id;
    private String name;
    private String salt;
    private String password;
    private Date createdAt;
}
