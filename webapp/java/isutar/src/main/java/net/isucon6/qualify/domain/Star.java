package net.isucon6.qualify.domain;

import java.util.Date;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class Star {
    private Long id;
    private String keyword;
    private String userName;
    private Date createdAt;

    public Star(String keyword, String userName) {
        this.keyword = keyword;
        this.userName = userName;
    }
}
