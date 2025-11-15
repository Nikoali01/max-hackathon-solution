package techaas.max_uni.uni_back.dao.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import techaas.max_uni.uni_back.dao.entity.Courses;

@Repository
public interface CoursesRepository extends JpaRepository<Courses, Long> {
}
