package techaas.max_uni.uni_back.service;

import techaas.max_uni.uni_back.model.rest.DayScheduleResponse;

import java.time.LocalDate;
import java.util.List;

public interface ScheduleService {

    List<DayScheduleResponse> constructDayScheduleForStudent(Long maxId, LocalDate date);
    List<DayScheduleResponse> constructProfessorDaySchedule(Long maxId, LocalDate date);
}
