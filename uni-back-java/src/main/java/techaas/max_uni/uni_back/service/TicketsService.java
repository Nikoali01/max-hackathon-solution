package techaas.max_uni.uni_back.service;

import org.springframework.stereotype.Service;
import techaas.max_uni.uni_back.dao.entity.Tickets;

import java.util.List;

public interface TicketsService {

    void createTicket(Tickets ticket);
    Tickets getTicket(String id);
    List<Tickets> getUserTickets(Long maxId);
}
