package techaas.max_uni.uni_back.service;

import techaas.max_uni.uni_back.dao.entity.Users;

public interface UsersService {

    Users registerUser(Users user) throws Exception;
    void validateUserCode(Long userId, String code) throws Exception;
    void deleteAllUsers();
    Users getUserInformation(Long maxId);
}
