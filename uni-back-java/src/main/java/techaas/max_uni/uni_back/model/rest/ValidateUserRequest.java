package techaas.max_uni.uni_back.model.rest;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ValidateUserRequest {

    private Long maxId;
    private String code;
}
