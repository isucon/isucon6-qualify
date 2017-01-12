package net.isucon6.qualify.exception.handler;

import java.io.IOException;

import org.springframework.http.HttpStatus;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.web.client.DefaultResponseErrorHandler;

/**
 * クライアントエラーを自分でハンドルするための、
 * 独自のRestTemplate用エラーハンドラ
 */
public class IsutarExceptionHandler extends DefaultResponseErrorHandler {
    @Override
    public boolean hasError(ClientHttpResponse res) throws IOException {
        HttpStatus.Series series = res.getStatusCode().series();
        return (HttpStatus.Series.SERVER_ERROR.equals(series));
    }
}
