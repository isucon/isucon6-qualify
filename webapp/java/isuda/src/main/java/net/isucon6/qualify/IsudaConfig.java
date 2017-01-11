package net.isucon6.qualify;

import net.isucon6.qualify.advice.AuthenticateInterceptor;
import net.isucon6.qualify.advice.SetNameInterceptor;
import org.modelmapper.ModelMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurerAdapter;

@Configuration
public class IsudaConfig extends WebMvcConfigurerAdapter {
    @Autowired
    private SetNameInterceptor setNameInterceptor;
    @Autowired
    private AuthenticateInterceptor authenticateInterceptor;

    @Bean
    public ModelMapper modelMapper() {
        return new ModelMapper();
    }

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(setNameInterceptor).addPathPatterns("/**");
        registry.addInterceptor(authenticateInterceptor).addPathPatterns("/**");
    }
}
