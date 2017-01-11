package net.isucon6.qualify.domain;

import java.util.Date;

import lombok.Data;

@Data
public class Star {
    private Long id;
    private String keyword;
    private String userName;
    private Date createdAt;
}
