package techaas.max_uni.uni_back.dao.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import techaas.max_uni.uni_back.dao.entity.Tickets;
import techaas.max_uni.uni_back.dao.entity.Users;

import java.util.List;

@Repository
public interface TicketsRepository extends JpaRepository<Tickets, String> {

    List<Tickets> findTicketsByUser(Users user);
}
