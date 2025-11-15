package techaas.max_uni.uni_back.dao.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;
import techaas.max_uni.uni_back.dao.entity.AllowedUsers;

@Repository
public interface AllowedUsersRepository extends JpaRepository<AllowedUsers, Long> {

    @Query(value = """
            select * from allowed_users
            where email = :email
            """, nativeQuery = true)
    AllowedUsers findByEmail(String email);
}
