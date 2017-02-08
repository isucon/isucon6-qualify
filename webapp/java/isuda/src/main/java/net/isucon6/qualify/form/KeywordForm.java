package net.isucon6.qualify.form;

import lombok.Data;
import org.hibernate.validator.constraints.NotBlank;

@Data
public class KeywordForm {
    @NotBlank
    private String keyword;
    @NotBlank
    private String description;
}
