package net.isucon6.qualify.form;

import lombok.Data;
import org.hibernate.validator.constraints.NotBlank;

@Data
public class AuthenticateForm {
    @NotBlank
    private String name;
    @NotBlank
    private String password;
}
