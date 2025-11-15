package techaas.max_uni.uni_back.dao.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import techaas.max_uni.uni_back.dao.entity.Lessons;

import java.time.LocalDate;
import java.util.Date;
import java.util.List;

@Repository
public interface LessonsRepository extends JpaRepository<Lessons, Long> {

    @Query(value = """
            SELECT *
            FROM public.lessons l
            WHERE l.course_id = (
                SELECT s.course_id
                FROM public.students s
                JOIN public.users u ON s.user_id = u.id
                WHERE u.max_id = :maxId
                LIMIT 1
            ) AND l.date_time >= current_date
                                     AND l.date_time < current_date + interval '1 day';
            """, nativeQuery = true)
    List<Lessons> findStudentsLessonsPerDay(@Param("maxId") Long maxId, @Param("date") LocalDate date);

    @Query(value = """
             SELECT *
             FROM public.lessons l
             WHERE l.professor_id = (select id from users u where u.max_id = :maxId)
             AND l.date_time >= current_date
             AND l.date_time < current_date + interval '1 day';
            """, nativeQuery = true)
    List<Lessons> findProfessorsLessonsPerDay(@Param("maxId") Long maxId, @Param("date") LocalDate date);
}
