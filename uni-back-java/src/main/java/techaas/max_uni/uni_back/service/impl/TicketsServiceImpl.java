package techaas.max_uni.uni_back.service.impl;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import techaas.max_uni.uni_back.dao.entity.Tickets;
import techaas.max_uni.uni_back.dao.repository.TicketsRepository;
import techaas.max_uni.uni_back.dao.repository.UsersRepository;
import techaas.max_uni.uni_back.service.TicketsService;

import java.util.List;

@Service
@RequiredArgsConstructor
public class TicketsServiceImpl implements TicketsService {

    private final TicketsRepository ticketsRepository;
    private final UsersRepository usersRepository;

    @Override
    public void createTicket(Tickets ticket) {
        ticketsRepository.save(ticket);
    }

    @Override
    public Tickets getTicket(String id) {
        return ticketsRepository.findById(id).orElse(null);
    }

    @Override
    public List<Tickets> getUserTickets(Long maxId) {
        return ticketsRepository.findTicketsByUser(usersRepository.findUsersById((maxId)).get(0));
    }
}
