package techaas.max_uni.uni_back.controller;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RestController;
import techaas.max_uni.uni_back.dao.entity.Students;
import techaas.max_uni.uni_back.dao.repository.StudentsRepository;

@RestController
@RequiredArgsConstructor
public class MaxController {

    private final StudentsRepository studentsRepository;

    @GetMapping("/students/{studentId}")
    public ResponseEntity<Students> getMax(@PathVariable long studentId) {
        return ResponseEntity.ok(studentsRepository.findById(studentId));
    }
}
