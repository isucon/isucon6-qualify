package net.isucon6.qualify.domain;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class Keyword {
    private Long authorId;
    private String keyword;
    private String description;
    public Keyword(Long authorId, String keyword, String description) {
        this.authorId = authorId;
        this.keyword = keyword;
        this.description = description;
    }
}
