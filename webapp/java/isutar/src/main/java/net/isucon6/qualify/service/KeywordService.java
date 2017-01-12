package net.isucon6.qualify.service;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestOperations;

@Service
public class KeywordService {
    @Autowired
    private RestOperations isudaRestOperations;

    public boolean exists(String keyword) {
        // TODO: isudaを呼び出して、404が返ってきたらfalse
        ResponseEntity res = isudaRestOperations.getForEntity("/keyword/"/*TODO: url encode keyword*/, String.class);
        if (res.getStatusCode() == HttpStatus.NOT_FOUND) return false;

        return true;
    }
}
