package net.isucon6.qualify.domain;

import java.util.Date;

import lombok.Data;

@Data
public class Entry {
    Long id;
    Long authorId;
    String keyword;
    String description;
    Date updatedAt;
    Date createdAt;
}
