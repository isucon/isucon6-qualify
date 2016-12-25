package net.isucon6.qualify.domain;

import java.util.Date;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Star {
    private long id;
    private String keyword;
    private String userName;
    private Date createdAt;
}
