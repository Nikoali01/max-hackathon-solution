package techaas.max_uni.uni_back.dao.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
@Entity
@Table(name = "users")
public class Users {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String name;

    @Column(nullable = false)
    private String surname;

    private String patronymic;

    @Column(name = "max_id")
    private Long maxId;

    @Column(unique = true, nullable = false)
    private String email;

    private Integer age;

    private String role;

    @Column(name = "generated_code")
    private String generatedCode;

    private boolean verified;
}
