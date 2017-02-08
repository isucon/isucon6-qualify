package net.isucon6.qualify.response;

import java.util.List;

import lombok.Data;

@Data
public class StarResponse {
    private List<Star> stars;

    @Data
    public static class Star {
        private Long id;
        private String keyword;
        private String userName;
        private String createdAt;
    }
}
