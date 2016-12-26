package net.isucon6.qualify.dto;

import java.util.List;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class EntryDto {
    /* Entry */
    private Long id;
    private Long authorId;
    private String keyword;
    private String html;
    /* Star */
    private List<String> stars;
}
