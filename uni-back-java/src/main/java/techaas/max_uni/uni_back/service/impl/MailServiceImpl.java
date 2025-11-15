package techaas.max_uni.uni_back.service.impl;

import lombok.RequiredArgsConstructor;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.MimeMessageHelper;
import org.springframework.stereotype.Service;

import jakarta.mail.internet.MimeMessage;
import techaas.max_uni.uni_back.service.MailService;

import java.nio.charset.StandardCharsets;

@Service
@RequiredArgsConstructor
public class MailServiceImpl implements MailService {

    private static final String subjectRegistration = "Код для регистрации на Max-Uni";

    private final JavaMailSender mailSender;

    public void sendRegistrationCode(String to, String code) throws Exception {
        MimeMessage message = mailSender.createMimeMessage();

        MimeMessageHelper helper = new MimeMessageHelper(message, true, StandardCharsets.UTF_8.name());
        helper.setTo(to);
        helper.setSubject(subjectRegistration);

        helper.setText("Ваш код для верифицкации: " + code);

        mailSender.send(message);
    }
}
