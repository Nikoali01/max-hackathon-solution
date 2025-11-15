package techaas.max_uni.uni_back.service;

public interface MailService {

    void sendRegistrationCode(String email, String code) throws Exception;
}
