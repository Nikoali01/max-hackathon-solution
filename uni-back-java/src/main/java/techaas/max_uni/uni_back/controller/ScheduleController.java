package techaas.max_uni.uni_back.controller;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import techaas.max_uni.uni_back.service.ScheduleService;
import techaas.max_uni.uni_back.model.rest.DayScheduleResponse;

import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/schedule")
@RequiredArgsConstructor
public class ScheduleController {

    private final ScheduleService scheduleService;

    @GetMapping("/student/{maxId}")
    public ResponseEntity<List<DayScheduleResponse>> getDaySchedule(@PathVariable("maxId") Long maxId) {
        return ResponseEntity.ok(scheduleService.constructDayScheduleForStudent(maxId, LocalDate.now()));
    }

    @GetMapping("/professor/{maxId}")
    public ResponseEntity<List<DayScheduleResponse>> getProfessorSchedule(@PathVariable("maxId") Long maxId) {
        return ResponseEntity.ok(scheduleService.constructProfessorDaySchedule(maxId, LocalDate.now()));
    }
}
