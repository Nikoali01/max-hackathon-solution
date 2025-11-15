package techaas.max_uni.uni_back.dao.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import techaas.max_uni.uni_back.dao.entity.News;
import techaas.max_uni.uni_back.dao.entity.Students;

import java.util.Collection;
import java.util.List;

@Repository
public interface NewsRepository extends JpaRepository<News, Long> {

    @Query(value = """
                select *
                from public.news
                order by date desc
                limit :limit
            """, nativeQuery = true)
    List<News> findLast10News(@Param("limit") int limit);

    @Modifying
    @Query(value = """
            update public.news
            set text = :text
            where id = :id
            """, nativeQuery = true)
    void editText(@Param("text") String text, @Param("id") Long id);
}
