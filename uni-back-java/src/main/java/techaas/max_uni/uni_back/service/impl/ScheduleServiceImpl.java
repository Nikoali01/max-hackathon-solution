package techaas.max_uni.uni_back.service.impl;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import techaas.max_uni.uni_back.dao.entity.Lessons;
import techaas.max_uni.uni_back.dao.repository.LessonsRepository;
import techaas.max_uni.uni_back.model.rest.DayScheduleResponse;
import techaas.max_uni.uni_back.service.ScheduleService;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

@Component
@RequiredArgsConstructor
public class ScheduleServiceImpl implements ScheduleService {

    private final LessonsRepository lessonsRepository;

    @Override
    public List<DayScheduleResponse> constructDayScheduleForStudent(Long maxId, LocalDate date) {
        var scheduleInfo = lessonsRepository.findStudentsLessonsPerDay(maxId, date);
        List<DayScheduleResponse> daySchedule = new ArrayList<>(List.of());
        for (Lessons lesson : scheduleInfo) {
            var professor = lesson.getProfessor();
            daySchedule.add(new DayScheduleResponse(lesson.getDateTime().toLocalTime().withSecond(0), lesson.getLessonName(), "проф. " + professor.getName() + " " + professor.getSurname(), lesson.getPlace(), lesson.getDescription()));
        }
        return daySchedule;
    }

    @Override
    public List<DayScheduleResponse> constructProfessorDaySchedule(Long maxId, LocalDate date) {
        var scheduleInfo = lessonsRepository.findProfessorsLessonsPerDay(maxId, date);
        List<DayScheduleResponse> daySchedule = new ArrayList<>(List.of());
        for (Lessons lesson : scheduleInfo) {
            var professor = lesson.getProfessor();
            daySchedule.add(new DayScheduleResponse(lesson.getDateTime().toLocalTime().withSecond(0), lesson.getLessonName(), "проф. " + professor.getName() + " " + professor.getSurname(), lesson.getPlace(), lesson.getDescription()));
        }
        return daySchedule;
    }
}
