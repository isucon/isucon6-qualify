package net.isucon6.qualify;

import org.modelmapper.ModelMapper;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class IsudaConfig {
    @Bean
    public ModelMapper modelMapper() {
        return new ModelMapper();
    }
}
