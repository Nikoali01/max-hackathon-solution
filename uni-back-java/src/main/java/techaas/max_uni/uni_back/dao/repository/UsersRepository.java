package techaas.max_uni.uni_back.dao.repository;

import jakarta.transaction.Transactional;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;
import techaas.max_uni.uni_back.dao.entity.Users;

import java.util.List;

@Repository
public interface UsersRepository extends JpaRepository<Users, Long> {

    @Modifying
    @Transactional
    @Query(value = """
            update public.users
            set generated_code = :code
            where email = :email
            """, nativeQuery = true)
    void updateGeneratedCodeByEmail(String email, String code);

    Users findByMaxId(Long maxId);

    List<Users> findUsersById(Long id);
}
