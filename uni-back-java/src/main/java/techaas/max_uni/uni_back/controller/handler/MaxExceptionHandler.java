package techaas.max_uni.uni_back.controller.handler;

import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;
import techaas.max_uni.uni_back.exception.CustomException;

@ControllerAdvice
@Slf4j
public class MaxExceptionHandler {

    @ExceptionHandler
    public ResponseEntity<String> exceptionHandler(Exception e) {
        throw new CustomException(e.getMessage());
    }
}
