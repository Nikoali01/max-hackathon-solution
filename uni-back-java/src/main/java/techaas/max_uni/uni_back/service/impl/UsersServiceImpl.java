package techaas.max_uni.uni_back.service.impl;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import techaas.max_uni.uni_back.dao.entity.Lessons;
import techaas.max_uni.uni_back.dao.entity.Students;
import techaas.max_uni.uni_back.dao.entity.Users;
import techaas.max_uni.uni_back.dao.repository.*;
import techaas.max_uni.uni_back.service.UsersService;
import techaas.max_uni.uni_back.util.AesEncryptionService;

import java.time.LocalDate;
import java.util.Random;

@Service
@RequiredArgsConstructor
public class UsersServiceImpl implements UsersService {

    private final MailServiceImpl mailService;
    private final UsersRepository usersRepository;
    private final AesEncryptionService encryptionService;
    private final AllowedUsersRepository allowedUsersRepository;
    private final LessonsRepository lessonsRepository;
    private final CoursesRepository coursesRepository;
    private final StudentsRepository studentsRepository;

    private Random random = new Random();

    @Override
    public Users registerUser(Users user) throws Exception {
        var allowedUsers = allowedUsersRepository.findByEmail(user.getEmail());
        if (allowedUsers != null && user.getEmail().equals(allowedUsers.getEmail())) {
            user.setRole(allowedUsers.getRole());
        } else {
            user.setRole("applicant");
        }
        usersRepository.save(user);

        if (user.getRole().equals("student")) {
            var course = coursesRepository.getById(random.nextLong(10));

            var student = new Students();
            student.setCourse(course);
            student.setUser(user);
            studentsRepository.save(student);

            var firstLesson = new Lessons();
            firstLesson.setCourse(course);
            firstLesson.setLessonName("Программирование");
            firstLesson.setPlace("Корпус Б, ауд. 115");
            firstLesson.setDescription("Практика по Go. Подготовьте вопросы по goroutines.");
            firstLesson.setDateTime(LocalDate.now().atTime(3, 50));
            firstLesson.setProfessor(usersRepository.getById(random.nextLong(2, 8)));
            lessonsRepository.save(firstLesson);

            var secondLesson = new Lessons();
            secondLesson.setCourse(course);
            secondLesson.setLessonName("Мат. анализ");
            secondLesson.setPlace("Корпус А, ауд. 302");
            secondLesson.setDescription("Лекция. Возьмите тетрадь и калькулятор.");
            secondLesson.setDateTime(LocalDate.now().atTime(1, 20));
            secondLesson.setProfessor(usersRepository.getById(random.nextLong(2, 8)));
            lessonsRepository.save(secondLesson);
        }

        var code = encryptionService.generate6DigitCode();
        usersRepository.updateGeneratedCodeByEmail(user.getEmail(), encryptionService.encrypt(code));
        mailService.sendRegistrationCode(user.getEmail(), code);
        return user;
    }

    @Override
    public void validateUserCode(Long maxId, String code) throws Exception {
        Users user = usersRepository.findByMaxId(maxId);
        String decryptedCode = encryptionService.decrypt(user.getGeneratedCode());
        if (decryptedCode.equals(code)) {
            user.setVerified(true);
            user.setGeneratedCode(null);
            usersRepository.save(user);
        } else {
            throw new RuntimeException("Invalid user code");
        }
    }

    @Override
    public void deleteAllUsers() {
        usersRepository.deleteAll();
    }

    @Override
    public Users getUserInformation(Long maxId) {
        return usersRepository.findByMaxId(maxId);
    }
}
