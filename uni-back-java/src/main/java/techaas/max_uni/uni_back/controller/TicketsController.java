package techaas.max_uni.uni_back.controller;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import techaas.max_uni.uni_back.dao.entity.Tickets;
import techaas.max_uni.uni_back.dao.repository.TicketsRepository;
import techaas.max_uni.uni_back.service.TicketsService;

@RestController
@RequestMapping("/tickets")
@RequiredArgsConstructor
public class TicketsController {

    private TicketsService ticketsService;

    @GetMapping("/ticket/{ticketId}")
    public ResponseEntity<?> getTicket(@PathVariable String ticketId) {
        return ResponseEntity.ok(ticketsService.getTicket(ticketId));
    }

    @GetMapping("/user/{maxId}")
    public ResponseEntity<?> getTicketByUser(@PathVariable Long maxId) {
        return ResponseEntity.ok(ticketsService.getUserTickets(maxId));
    }

    @PostMapping("/save")
    public ResponseEntity<?> saveTicket(@RequestBody Tickets ticket) {
        ticketsService.createTicket(ticket);
        return ResponseEntity.ok().build();
    }
}
