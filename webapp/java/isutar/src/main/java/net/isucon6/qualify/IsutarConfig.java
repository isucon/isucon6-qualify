package net.isucon6.qualify;

import net.isucon6.qualify.exception.handler.IsutarExceptionHandler;
import org.modelmapper.ModelMapper;
import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestOperations;

@Configuration
public class IsutarConfig {
    @Bean
    public ModelMapper modelMapper() {
        return new ModelMapper();
    }

    @Bean
    public RestOperations isudaRestOperations(RestTemplateBuilder builder) {
        return builder.rootUri("http://localhost:5000")
                .errorHandler(new IsutarExceptionHandler()).build();
    }
}
