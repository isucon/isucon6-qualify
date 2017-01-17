package net.isucon6.qualify.service;

import java.util.HashMap;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestOperations;

@Service
public class SpamService {
    @Autowired
    private RestOperations isupamRestOperation;
    private Logger log = LoggerFactory.getLogger(SpamService.class);

    public boolean isSpam(String text) {
        Map<String, String> params = new HashMap<String, String>() {{
            put("content", text);
        }};
        ResponseEntity<Map> res = isupamRestOperation.postForEntity("/", params, Map.class);
        return !Boolean.valueOf(String.valueOf(res.getBody().get("valid")));
    }
}
