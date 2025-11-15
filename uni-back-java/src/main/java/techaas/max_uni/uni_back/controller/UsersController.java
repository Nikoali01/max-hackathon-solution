package techaas.max_uni.uni_back.controller;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import techaas.max_uni.uni_back.dao.entity.Users;
import techaas.max_uni.uni_back.dao.repository.UsersRepository;
import techaas.max_uni.uni_back.model.rest.ValidateUserRequest;
import techaas.max_uni.uni_back.service.UsersService;

@RestController
@RequestMapping("/users")
@RequiredArgsConstructor
public class UsersController {

    private final UsersService usersService;

    @GetMapping("/{maxId}")
    public ResponseEntity<Users> getUser(@PathVariable Long maxId) {
        return ResponseEntity.ok(usersService.getUserInformation(maxId));
    }

    @PostMapping(value = "/register")
    public ResponseEntity<Users> createUser(@RequestBody Users user) throws Exception {
        return ResponseEntity.ok(usersService.registerUser(user));
    }

    @PostMapping(value = "/validate")
    public ResponseEntity<?> validateUserCode(@RequestBody ValidateUserRequest request) throws Exception {
        usersService.validateUserCode(request.getMaxId(), request.getCode());
        return ResponseEntity.ok().build();
    }

    @PostMapping("/clear")
    public ResponseEntity<?> deleteAllUsers() {
        usersService.deleteAllUsers();
        return ResponseEntity.ok().build();
    }
}
