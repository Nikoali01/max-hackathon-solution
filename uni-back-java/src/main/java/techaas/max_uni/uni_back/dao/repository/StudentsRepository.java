package techaas.max_uni.uni_back.dao.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import techaas.max_uni.uni_back.dao.entity.Students;

@Repository
public interface StudentsRepository extends JpaRepository<Students, Long> {

    @Query(value = """
                select *
                from public.students s
                where id = :id
            """, nativeQuery = true)
    Students findById(@Param("id") long id);

    @Query(value = """
                select course_id
                from public.students s
                where id = :id
            """, nativeQuery = true)
    Students getStudentsCourse(@Param("id") long id);
}
